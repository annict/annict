# frozen_string_literal: true

module Db
  class SearchesController < Db::ApplicationController
    def show
      @works = @search.works.order(id: :desc)
      @people = @search.people.order(id: :desc)
      @organizations = @search.organizations.order(id: :desc)
    end
  end
end
