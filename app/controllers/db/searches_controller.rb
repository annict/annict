module Db
  class SearchesController < Db::ApplicationController
    def show
      @works = @search.works
      @people = @search.people
    end
  end
end
