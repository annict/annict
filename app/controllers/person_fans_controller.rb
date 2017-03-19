# frozen_string_literal: true

class PersonFansController < ApplicationController
  before_action :load_i18n, only: %i(index)

  def index(person_id)
    @person = Person.published.find(person_id)
    @fan_users = @person.users.order("favorite_people.id DESC")
  end

  private

  def load_i18n
    keys = {
      "messages.components.favorite_button.add_to_favorites": nil,
      "messages.components.favorite_button.added_to_favorites": nil,
    }

    load_i18n_into_gon keys
  end
end
