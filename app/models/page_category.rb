# frozen_string_literal: true

class PageCategory
  NAMES = %i[
    cast_list
    character_fan_list
    collection_list
    episode
    episode_list
    favorite_character_list
    favorite_organization_list
    favorite_person_list
    followee_list
    follower_list
    home
    library
    newest_work_list
    organization_fan_list
    person_fan_list
    popular_work_list
    profile
    record
    record_edit
    record_list
    related_work_list
    search
    seasonal_work_list
    slot_list
    staff_list
    track
    video_list
    welcome
    work
    work_info
    work_record_list
  ].freeze

  NAMES.each do |name|
    const_set(name.upcase, name.to_s.dasherize)
  end

  def self.all
    NAMES.map { |name| const_get name.upcase }
  end
end
