# frozen_string_literal: true

class EpisodeRecordForm < ApplicationForm
  attr_accessor :episode, :oauth_application, :record
  attr_reader :comment, :rating, :share_to_twitter

  validates :comment, length: {maximum: 1}
  validates :rating, inclusion: {in: Record::RATING_STATES.map(&:to_s)}, allow_nil: true

  def comment=(comment)
    @comment = comment&.strip
  end

  def rating=(rating)
    @rating = rating&.downcase.presence
  end

  def share_to_twitter=(share_to_twitter)
    @share_to_twitter = ActiveRecord::Type::Boolean.new.serialize(share_to_twitter)
  end

  def persisted?
    !record.nil?
  end

  def unique_id
    @unique_id ||= SecureRandom.uuid
  end
end
