# frozen_string_literal: true

module Api
  module Internal
    class SearchesController < Api::Internal::ApplicationController
      def show
        search = SearchService.new(params[:q])
        @works = search.works.order(id: :desc).limit(5)
        @people = search.people.order(id: :desc).limit(5)
        @organizations = search.organizations.order(id: :desc).limit(5)
        @characters = search.characters.order(id: :desc).limit(5)
      end
    end
  end
end
