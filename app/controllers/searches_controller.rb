# frozen_string_literal: true

class SearchesController < ApplicationController
  def show
    @works, @characters, @people, @organizations = if params[:q]
      [
        @search.works.order(id: :desc),
        @search.characters.order(id: :desc),
        @search.people.order(id: :desc),
        @search.organizations.order(id: :desc)
      ]
    else
      [Work.none, Character.none, Person.none, Organization.none]
    end
    @view = select_view(params[:resource].presence || "work")
  end

  private

  def select_view(resource)
    resources = %w(work character person organization)
    return resource if resource.in?(resources)
    collection = [@works, @characters, @people, @organizations].
      select { |c| c.count.positive? }.
      sort_by(&:count).
      reverse.
      first
    return collection.model.name.downcase if collection.present?
    "work"
  end
end
