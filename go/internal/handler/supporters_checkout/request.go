package supporters_checkout

// CreateRequest はCheckoutセッション作成のリクエストパラメータ
type CreateRequest struct {
	Plan string
}

// Validate はリクエストパラメータのバリデーションを行います
func (r *CreateRequest) Validate() map[string]string {
	errors := make(map[string]string)
	if r.Plan != "monthly" && r.Plan != "yearly" {
		errors["plan"] = "invalid_plan"
	}
	return errors
}
