module Annict
  class Subdomain
    def self.matches?(request)
      request.subdomain == "api"
    end
  end
end
