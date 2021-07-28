# frozen_string_literal: true

class PageCategory
  NAMES = %i[
    anime
    anime_info
    anime_record_list
    cast_list
    character_fan_list
    episode
    episode_list
    favorite_character_list
    favorite_organization_list
    favorite_person_list
    followee_list
    follower_list
    home
    library
    newest_anime_list
    organization_fan_list
    person_fan_list
    popular_anime_list
    profile
    record
    record_edit
    record_list
    related_anime_list
    search
    seasonal_anime_list
    slot_list
    staff_list
    track
    video_list
    welcome
  ].freeze

  NAMES.each do |name|
    const_set(name.upcase, name.to_s.dasherize)
  end

  def self.all
    NAMES.map { |name| const_get name.upcase }
  end
end
