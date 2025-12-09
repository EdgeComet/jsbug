package server

import "github.com/user/jsbug/internal/types"

// enrichImagesWithSizes matches extracted images with network requests
// and populates the Size field from request data.
func enrichImagesWithSizes(images []types.Image, requests []types.NetworkRequest) {
	// Build lookup map: URL -> size
	sizeMap := make(map[string]int)
	for _, req := range requests {
		if req.Type == "Image" && req.Size > 0 {
			sizeMap[req.URL] = req.Size
		}
	}

	// Direct lookup - Src is already resolved to absolute URL
	for i := range images {
		if size, ok := sizeMap[images[i].Src]; ok {
			images[i].Size = size
		}
	}
}
