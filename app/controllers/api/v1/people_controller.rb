# frozen_string_literal: true

module Api
  module V1
    class PeopleController < Api::V1::ApplicationController
      before_action :prepare_params!, only: %i[index]

      def index
        @people = Person.only_kept
        @people = Api::V1::PersonIndexService.new(@people, @params).result
      end
    end
  end
end
