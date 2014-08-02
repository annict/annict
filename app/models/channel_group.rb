class ChannelGroup < ActiveRecord::Base
  has_many :channels

  scope :published, -> { where.not(sort_number: nil) }
end