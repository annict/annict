# typed: false
# frozen_string_literal: true

module Api
  module Internal
    class OrganizationsController < Api::Internal::ApplicationController
      def index
        q = params[:q]
        @organizations = if q
          Organization.where("name ILIKE ?", "%#{q}%").only_kept
        else
          Organization.none
        end
      end
    end
  end
end
