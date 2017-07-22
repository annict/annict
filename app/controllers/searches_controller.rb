# frozen_string_literal: true

class SearchesController < ApplicationController
  before_action :load_i18n, only: %i(show)

  def show(q: nil, resource: nil)
    @works, @characters, @people, @organizations = if q.present?
      [
        @search.works.order(id: :desc),
        @search.characters.order(id: :desc),
        @search.people.order(id: :desc),
        @search.organizations.order(id: :desc)
      ]
    else
      [Work.none, Character.none, Person.none, Organization.none]
    end
    @view = select_view resource

    return unless user_signed_in?

    gon.pageObject = render_jb "works/_list",
      user: current_user,
      works: @works
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

  def load_i18n
    keys = {
      "messages._components.collect_button_modal.added": nil,
      "messages._components.collect_button_modal.view_collection": nil
    }

    load_i18n_into_gon keys
  end
end
