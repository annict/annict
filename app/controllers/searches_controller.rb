# frozen_string_literal: true

class SearchesController < ApplicationController
  def show(q: nil)
    @works, @people, @organizations = if q.present?
      [
        @search.works.order(id: :desc).limit(20),
        @search.people.order(id: :desc).limit(20),
        @search.organizations.order(id: :desc).limit(20)
      ]
    else
      [Work.none, Person.none, Organization.none]
    end

    render layout: "v1/application"
  end
end
