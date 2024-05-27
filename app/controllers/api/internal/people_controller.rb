# typed: false
# frozen_string_literal: true

module Api
  module Internal
    class PeopleController < Api::Internal::ApplicationController
      def index
        @people = if params[:q].present?
          Person.where("name ILIKE ?", "%#{params[:q]}%").only_kept
        else
          Person.none
        end
      end
    end
  end
end
