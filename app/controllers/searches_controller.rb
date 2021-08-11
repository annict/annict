# frozen_string_literal: true

class SearchesController < ApplicationV6Controller
  def show
    set_page_category PageCategory::SEARCH

    @works, @characters, @people, @organizations = if params[:q]
      [
        @search.works.order(id: :desc),
        @search.characters.order(id: :desc),
        @search.people.order(id: :desc),
        @search.organizations.order(id: :desc)
      ]
    else
      [Anime.none, Character.none, Person.none, Organization.none]
    end
    @view = select_view(params[:resource].presence || "work")

    @resources, @partial_name = case @view
    when "work"
      [@works, "work_list"]
    when "character"
      [@characters, "character_list"]
    when "person"
      [@people, "person_list"]
    when "organization"
      [@organizations, "organization_list"]
    end
    @resources = @resources.page(params[:page])
  end

  private

  def select_view(resource)
    resources = %w[work character person organization]
    return resource if resource.in?(resources)
    collection = [@works, @characters, @people, @organizations]
      .select { |c| c.count.positive? }
      .max_by(&:count)

    return collection.model.name.downcase if collection.present?
    "work"
  end
end
