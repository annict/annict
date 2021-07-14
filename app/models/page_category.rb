# frozen_string_literal: true

class PageCategory
  NAMES = %i[
    anime
    anime_record_list
    character_fan_list
    episode
    episode_list
    favorite_character_list
    favorite_organization_list
    favorite_person_list
    follower_list
    following_list
    home
    library
    organization_fan_list
    person_fan_list
    profile
    record
    record_edit
    record_list
    search
    seasonal_anime_list
    slot_list
    track
    welcome
    work_list_newest
    work_list_popular
  ].freeze

  NAMES.each do |name|
    const_set(name.upcase, name.to_s.dasherize)
  end

  def self.all
    NAMES.map { |name| const_get name.upcase }
  end
end
