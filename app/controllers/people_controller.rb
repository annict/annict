# frozen_string_literal: true

class PeopleController < ApplicationV6Controller
  def show
    @person = Person.only_kept.find(params[:person_id])

    if @person.voice_actor?
      @casts_with_year = @person
        .casts
        .only_kept
        .joins(:anime)
        .where(works: {deleted_at: nil})
        .includes(:character, anime: :anime_image)
        .group_by { |cast| cast.anime.season_year.presence || 0 }
      @cast_years = @casts_with_year.keys.sort.reverse
    end

    if @person.staff?
      @staffs_with_year = @person
        .staffs
        .only_kept
        .joins(:anime)
        .where(works: {deleted_at: nil})
        .includes(anime: :anime_image)
        .group_by { |staff| staff.anime.season_year.presence || 0 }
      @staff_years = @staffs_with_year.keys.sort.reverse
    end

    @person_favorites = @person
      .person_favorites
      .includes(user: :profile)
      .joins(:user)
      .merge(User.only_kept)
      .order(id: :desc)
      .limit(8)
  end
end
