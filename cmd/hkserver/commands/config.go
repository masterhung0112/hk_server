package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/masterhung0112/hk_server/v5/config"
	"github.com/masterhung0112/hk_server/v5/model"
	"github.com/masterhung0112/hk_server/v5/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const noSettingsNamed = "unable to find a setting named: %s"

var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration",
}

var ConfigSetCmd = &cobra.Command{
	Use:     "set",
	Short:   "Set config setting",
	Long:    "Sets the value of a config setting by its name in dot notation. Accepts multiple values for array settings",
	Example: "config set SqlSettings.DriverName mysql",
	Args:    cobra.MinimumNArgs(2),
	RunE:    configSetCmdF,
}

func init() {
	// ConfigSubpathCmd.Flags().String("path", "", "Optional subpath; defaults to value in SiteURL")
	// ConfigResetCmd.Flags().Bool("confirm", false, "Confirm you really want to reset all configuration settings to its default value")
	// ConfigShowCmd.Flags().Bool("json", false, "Output the configuration as JSON.")

	ConfigCmd.AddCommand(
		// ValidateConfigCmd,
		// ConfigSubpathCmd,
		// ConfigGetCmd,
		// ConfigShowCmd,
		ConfigSetCmd,
		// MigrateConfigCmd,
		// ConfigResetCmd,
	)
	RootCmd.AddCommand(ConfigCmd)
}

func getConfigStore(command *cobra.Command) (*config.Store, error) {
	if err := utils.TranslationsPreInit(); err != nil {
		return nil, errors.Wrap(err, "failed to initialize i18n")
	}

	configStore, err := config.NewStore(getConfigDSN(command, config.GetEnvironment()), false, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize config store")
	}

	return configStore, nil
}

func configSetCmdF(command *cobra.Command, args []string) error {
	configStore, err := getConfigStore(command)
	if err != nil {
		return err
	}

	// args[0] -> holds the config setting that we want to change
	// args[1:] -> the new value of the config setting
	configSetting := args[0]
	newVal := args[1:]

	// create the function to update config
	oldConfig := configStore.Get()
	newConfig := configStore.Get()

	f := updateConfigValue(configSetting, newVal, oldConfig, newConfig)
	f(newConfig)

	// UpdateConfig above would have already fixed these invalid locales, but we check again
	// in the context of an explicit change to these parameters to avoid saving the fixed
	// settings in the first place.
	if changed := config.FixInvalidLocales(newConfig); changed {
		return errors.New("Invalid locale configuration")
	}

	if _, errSet := configStore.Set(newConfig); errSet != nil {
		return errors.Wrap(errSet, "failed to set config")
	}

	/*
		Uncomment when CI unit test fail resolved.

		a, errInit := InitDBCommandContextCobra(command)
		if errInit == nil {
			auditRec := a.MakeAuditRecord("configSet", audit.Success)
			auditRec.AddMeta("setting", configSetting)
			auditRec.AddMeta("new_value", newVal)
			a.LogAuditRec(auditRec, nil)
			a.Srv().Shutdown()
		}
	*/

	return nil
}

func updateConfigValue(configSetting string, newVal []string, oldConfig, newConfig *model.Config) func(*model.Config) {
	return func(update *model.Config) {

		// convert config to map[string]interface
		configMap := configToMap(*oldConfig)

		// iterate through the map and update the value or print an error and exit
		err := UpdateMap(configMap, strings.Split(configSetting, "."), newVal)
		if err != nil {
			fmt.Printf("%s\n", err)
			os.Exit(1)
		}

		// convert map to json
		bs, err := json.Marshal(configMap)
		if err != nil {
			fmt.Printf("Error while marshalling map to json %s\n", err)
			os.Exit(1)
		}

		// convert json to struct
		err = json.Unmarshal(bs, newConfig)
		if err != nil {
			fmt.Printf("Error while unmarshalling json to struct %s\n", err)
			os.Exit(1)
		}

		*update = *newConfig

	}
}

func UpdateMap(configMap map[string]interface{}, configSettings []string, newVal []string) error {
	res, ok := configMap[configSettings[0]]
	if !ok {
		return fmt.Errorf(noSettingsNamed, configSettings[0])
	}

	value := reflect.ValueOf(res)

	switch value.Kind() {

	case reflect.Map:
		// we can only change the value of a particular setting, not the whole map, return error
		if len(configSettings) == 1 {
			return errors.New("unable to set multiple settings at once")
		}
		simpleMap, ok := res.(map[string]interface{})
		if ok {
			return UpdateMap(simpleMap, configSettings[1:], newVal)
		}
		mapOfTheMap, ok := res.(map[string]map[string]interface{})
		if ok {
			convertedMap := make(map[string]interface{})
			for k, v := range mapOfTheMap {
				convertedMap[k] = v
			}
			return UpdateMap(convertedMap, configSettings[1:], newVal)
		}
		//TODO: Open
		// pluginStateMap, ok := res.(map[string]*model.PluginState)
		// if ok {
		// 	convertedMap := make(map[string]interface{})
		// 	for k, v := range pluginStateMap {
		// 		convertedMap[k] = v
		// 	}
		// 	return UpdateMap(convertedMap, configSettings[1:], newVal)
		// }
		return fmt.Errorf(noSettingsNamed, configSettings[1])

	case reflect.Int:
		if len(configSettings) == 1 {
			val, err := strconv.Atoi(newVal[0])
			if err != nil {
				return err
			}
			configMap[configSettings[0]] = val
			return nil
		}
		return fmt.Errorf(noSettingsNamed, configSettings[0])

	case reflect.Int64:
		if len(configSettings) == 1 {
			val, err := strconv.Atoi(newVal[0])
			if err != nil {
				return err
			}
			configMap[configSettings[0]] = int64(val)
			return nil
		}
		return fmt.Errorf(noSettingsNamed, configSettings[0])

	case reflect.Bool:
		if len(configSettings) == 1 {
			val, err := strconv.ParseBool(newVal[0])
			if err != nil {
				return err
			}
			configMap[configSettings[0]] = val
			return nil
		}
		return fmt.Errorf(noSettingsNamed, configSettings[0])

	case reflect.String:
		if len(configSettings) == 1 {
			configMap[configSettings[0]] = newVal[0]
			return nil
		}
		return fmt.Errorf(noSettingsNamed, configSettings[0])

	case reflect.Slice:
		if len(configSettings) == 1 {
			configMap[configSettings[0]] = newVal
			return nil
		}
		return fmt.Errorf(noSettingsNamed, configSettings[0])

		//TODO: Open
	// case reflect.Ptr:
	// 	state, ok := res.(*model.PluginState)
	// 	if !ok || len(configSettings) != 2 {
	// 		return errors.New("type not supported yet")
	// 	}
	// 	val, err := strconv.ParseBool(newVal[0])
	// 	if err != nil {
	// 		return err
	// 	}
	// 	state.Enable = val
	// 	return nil

	default:
		return errors.New("type not supported yet")
	}
}

// configToMap converts our config into a map
func configToMap(s interface{}) map[string]interface{} {
	return structToMap(s)
}
