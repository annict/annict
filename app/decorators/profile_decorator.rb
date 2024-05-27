# typed: false

module ProfileDecorator
  def shorten_url
    uri = URI(url)
    path = uri.path == "/" ? "" : uri.path
    "#{uri.host}#{path}"
  end
end
