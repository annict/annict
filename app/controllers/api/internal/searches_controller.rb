# frozen_string_literal: true

module Api
  module Internal
    class SearchesController < Api::ApplicationController
      def show(q: nil)
        @results = SearchService.new(q).all.to_a
      end
    end
  end
end
