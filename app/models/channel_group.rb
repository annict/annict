# == Schema Information
#
# Table name: channel_groups
#
#  id          :integer          not null, primary key
#  sc_chgid    :string(510)      not null
#  name        :string(510)      not null
#  sort_number :integer
#  created_at  :datetime
#  updated_at  :datetime
#
# Indexes
#
#  channel_groups_sc_chgid_key  (sc_chgid) UNIQUE
#

class ChannelGroup < ApplicationRecord
  has_many :channels

  scope :published, -> { where.not(sort_number: nil) }
end
