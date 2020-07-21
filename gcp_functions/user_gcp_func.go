package gcp_functions

import (
  "net/http"
)

/**
 * URL: /users
 */
func user_gcp_func(w http.ResponseWriter, r *http.Request) {

  // Initialize the server function
  // th := SetupHelper()//.InitBasic()
  // defer th.TearDown()

  // Call function to http get
  switch r.Method {
  case http.MethodGet:
    // getUsers
    break
  case http.MethodPost:
    // createUser
    break

  }
}