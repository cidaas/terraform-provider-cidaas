resource "cidaas_hosted_page" "sample_hosted_page" {
  hosted_page_group_name = "sample_hosted_page_group"
  default_locale         = "en"
  hosted_pages = [
    {
      hosted_page_id = "login"
      content        = "<h1>Login</h1>"
      locale         = "en"
      url            = "https://example.com/login"
    }
  ]
}
