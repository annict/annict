# frozen_string_literal: true

module Api
  module Internal
    class SearchesController < Api::Internal::ApplicationController
      def show(q: nil)
        search = SearchService.new(q)
        @works = search.works.order(id: :desc).limit(5)
        @people = search.people.order(id: :desc).limit(5)
        @organizations = search.organizations.order(id: :desc).limit(5)
      end
    end
  end
end
