package requests

type AddImagesRequest struct {
	Images []string `mapstructure:"images"`
}
