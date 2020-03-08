# frozen_string_literal: true

module API
  module V1
    class PeopleController < API::V1::ApplicationController
      before_action :prepare_params!, only: %i(index)

      def index
        @people = Person.without_deleted
        @people = API::V1::PersonIndexService.new(@people, @params).result
      end
    end
  end
end
