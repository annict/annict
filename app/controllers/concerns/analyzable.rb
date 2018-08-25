# frozen_string_literal: true

module Analyzable
  extend ActiveSupport::Concern

  included do
    def ga_client
      @ga_client ||= Annict::Analytics::Client.new(request, current_user)
    end
  end
end
