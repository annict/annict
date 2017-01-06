# frozen_string_literal: true
# == Schema Information
#
# Table name: channels
#
#  id               :integer          not null, primary key
#  channel_group_id :integer          not null
#  sc_chid          :integer          not null
#  name             :string           not null
#  published        :boolean          default(TRUE), not null
#  created_at       :datetime
#  updated_at       :datetime
#
# Indexes
#
#  channels_channel_group_id_idx  (channel_group_id)
#  channels_sc_chid_key           (sc_chid) UNIQUE
#

class ChannelsController < ApplicationController
  before_action :authenticate_user!

  def index
    @channel_groups = ChannelGroup.published.order(:sort_number)
  end
end
