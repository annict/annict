# frozen_string_literal: true

module Db
  class SearchesController < Db::ApplicationController
    def show
      @works = @search.works.
        includes(:season, :item).
        order(id: :desc)
      @people = @search.people.order(id: :desc)
      @organizations = @search.organizations.order(id: :desc)
      @characters = @search.characters.order(id: :desc)
    end
  end
end
