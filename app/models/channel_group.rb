# frozen_string_literal: true
# == Schema Information
#
# Table name: channel_groups
#
#  id          :integer          not null, primary key
#  deleted_at  :datetime
#  name        :string(510)      not null
#  sc_chgid    :string(510)
#  sort_number :integer          default(0), not null
#  created_at  :datetime
#  updated_at  :datetime
#
# Indexes
#
#  channel_groups_sc_chgid_key  (sc_chgid) UNIQUE
#

class ChannelGroup < ApplicationRecord
  include SoftDeletable

  has_many :channels
end
