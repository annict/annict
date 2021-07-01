# frozen_string_literal: true

class PersonFansController < ApplicationV6Controller
  def index
    set_page_category PageCategory::PERSON_FAN_LIST

    @person = Person.only_kept.find(params[:person_id])
    @person_favorites = @person
      .person_favorites
      .joins(:user)
      .merge(User.only_kept)
      .order(watched_works_count: :desc)
  end
end
