# frozen_string_literal: true

# == Schema Information
#
# Table name: channel_groups
#
#  id             :bigint           not null, primary key
#  deleted_at     :datetime
#  name           :string(510)      not null
#  sc_chgid       :string(510)
#  sort_number    :integer          default(0), not null
#  unpublished_at :datetime
#  created_at     :timestamptz
#  updated_at     :timestamptz
#
# Indexes
#
#  channel_groups_sc_chgid_key             (sc_chgid) UNIQUE
#  index_channel_groups_on_unpublished_at  (unpublished_at)
#

class ChannelGroup < ApplicationRecord
  include Unpublishable

  has_many :channels
end
