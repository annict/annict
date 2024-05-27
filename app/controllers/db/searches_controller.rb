# typed: false
# frozen_string_literal: true

module Db
  class SearchesController < Db::ApplicationController
    def show
      @results = {
        series: @search.series_list.order(id: :desc).limit(100),
        work: @search.works.preload(:work_image).order(id: :desc).limit(100),
        person: @search.people.order(id: :desc).limit(100),
        organization: @search.organizations.order(id: :desc).limit(100),
        character: @search.characters.preload(:series).order(id: :desc).limit(100)
      }
    end
  end
end
