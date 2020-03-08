# frozen_string_literal: true

module API
  module Internal
    class PeopleController < API::Internal::ApplicationController
      def index
        @people = if params[:q].present?
          Person.where("name ILIKE ?", "%#{params[:q]}%").without_deleted
        else
          Person.none
        end
      end
    end
  end
end
