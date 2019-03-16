# frozen_string_literal: true
# == Schema Information
#
# Table name: channel_groups
#
#  id             :integer          not null, primary key
#  sc_chgid       :string(510)
#  name           :string(510)      not null
#  sort_number    :integer          default(0), not null
#  created_at     :datetime
#  updated_at     :datetime
#  unpublished_at :datetime
#
# Indexes
#
#  channel_groups_sc_chgid_key  (sc_chgid) UNIQUE
#

class ChannelGroup < ApplicationRecord
  include Publishable

  has_many :channels
end
