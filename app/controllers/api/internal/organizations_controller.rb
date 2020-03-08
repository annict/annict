# frozen_string_literal: true

module API
  module Internal
    class OrganizationsController < API::Internal::ApplicationController
      def index
        q = params[:q]
        @organizations = if q
          Organization.where("name ILIKE ?", "%#{q}%").without_deleted
        else
          Organization.none
        end
      end
    end
  end
end
