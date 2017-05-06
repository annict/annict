# frozen_string_literal: true

module Db
  class SearchesController < Db::ApplicationController
    def show
      @series_list = @search.series_list.order(id: :desc)
      @works = @search.works.
        order(id: :desc)
      @people = @search.people.order(id: :desc)
      @organizations = @search.organizations.order(id: :desc)
      @characters = @search.characters.order(id: :desc)
    end
  end
end
