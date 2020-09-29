# frozen_string_literal: true

class PageCategory
  NAMES = %i(
    character_fan_list
    episode
    episode_list
    favorite_character_list
    favorite_organization_list
    favorite_person_list
    follower_list
    following_list
    guest_home
    library
    organization_fan_list
    person_fan_list
    profile
    record
    record_edit
    record_list
    search
    slot_list
    track
    user_home
    user_work_tag
    work
    work_list_newest
    work_list_popular
    work_list_season
    work_record_list
  ).freeze

  NAMES.each do |name|
    const_set(name.upcase, name)
  end

  def self.all
    NAMES.map { |name| const_get name.upcase }
  end
end
