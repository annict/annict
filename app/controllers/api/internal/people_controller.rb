# frozen_string_literal: true

module Api
  module Internal
    class PeopleController < Api::Internal::ApplicationController
      def index(q: nil)
        @people = if q.present?
          Person.where("name ILIKE ?", "%#{q}%").published
        else
          Person.none
        end
      end
    end
  end
end
