# frozen_string_literal: true

module Doorkeeper
  class AccessTokenDecorator < ApplicationDecorator
    def local_scopes
      scopes.to_a.map do |scope|
        I18n.t("doorkeeper.scopes.#{scope}")
      end.join(", ")
    end
  end
end
