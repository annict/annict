# frozen_string_literal: true

module Api
  module Internal
    class OrganizationsController < Api::Internal::ApplicationController
      def index(q: nil)
        @organizations = if q.present?
          Organization.where("name ILIKE ?", "%#{q}%").published
        else
          Organization.none
        end
      end
    end
  end
end
