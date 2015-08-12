# == Schema Information
#
# Table name: channel_groups
#
#  id          :integer          not null, primary key
#  sc_chgid    :string           not null
#  name        :string           not null
#  sort_number :integer
#  created_at  :datetime
#  updated_at  :datetime
#
# Indexes
#
#  index_channel_groups_on_sc_chgid  (sc_chgid) UNIQUE
#

class ChannelGroup < ActiveRecord::Base
  has_many :channels

  scope :published, -> { where.not(sort_number: nil) }
end
