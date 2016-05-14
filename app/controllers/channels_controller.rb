# frozen_string_literal: true
# == Schema Information
#
# Table name: channels
#
#  id               :integer          not null, primary key
#  channel_group_id :integer          not null
#  sc_chid          :integer          not null
#  name             :string           not null
#  created_at       :datetime
#  updated_at       :datetime
#  published        :boolean          default(TRUE), not null
#
# Indexes
#
#  index_channels_on_published  (published)
#  index_channels_on_sc_chid    (sc_chid) UNIQUE
#

class ChannelsController < ApplicationController
  before_action :authenticate_user!

  def index
    @channel_groups = ChannelGroup.published.order(:sort_number)

    render layout: "v1/application"
  end
end
