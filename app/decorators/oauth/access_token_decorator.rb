# frozen_string_literal: true

module Oauth
  module AccessTokenDecorator
    def local_scopes
      scopes.to_a.map { |scope|
        I18n.t("doorkeeper.scopes.#{scope}")
      }.join(", ")
    end
  end
end
