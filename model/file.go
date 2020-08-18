package model

var (
	IMAGE_EXTENSIONS = [7]string{".jpg", ".jpeg", ".gif", ".bmp", ".png", ".tiff", "tif"}
	IMAGE_MIME_TYPES = map[string]string{".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".gif": "image/gif", ".bmp": "image/bmp", ".png": "image/png", ".tiff": "image/tiff", ".tif": "image/tif"}
)