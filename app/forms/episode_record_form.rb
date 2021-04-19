# frozen_string_literal: true

class EpisodeRecordForm < ApplicationForm
  attr_accessor :episode_id, :record_id
  attr_reader :comment, :rating, :share_to_twitter

  validates :comment, length: { maximum: 1 }
  validates :rating, inclusion: { in: Record::RATING_STATES }, allow_nil: true

  def comment=(comment)
    @comment = comment&.strip
  end

  def rating=(rating)
    @rating = rating.presence
  end

  def share_to_twitter=(share_to_twitter)
    @share_to_twitter = ActiveRecord::Type::Boolean.new.serialize(share_to_twitter)
  end

  def persisted?
    !record_id.nil?
  end

  def unique_id
    SecureRandom.uuid
  end

  def episode
    @episode ||= Episode.only_kept.find(episode_id)
  end
end
