# frozen_string_literal: true
# == Schema Information
#
# Table name: channel_works
#
#  id         :integer          not null, primary key
#  user_id    :integer          not null
#  work_id    :integer          not null
#  channel_id :integer          not null
#  created_at :datetime
#  updated_at :datetime
#
# Indexes
#
#  channel_works_channel_id_idx                  (channel_id)
#  channel_works_user_id_idx                     (user_id)
#  channel_works_user_id_work_id_channel_id_key  (user_id,work_id,channel_id) UNIQUE
#  channel_works_work_id_idx                     (work_id)
#

class ChannelWorksController < ApplicationController
  before_action :authenticate_user!

  def index
    @works = current_user.
      works.
      wanna_watch_and_watching.
      published.
      slot_registered.
      includes(:episodes, :work_image).
      order_by_season(:desc)
  end
end
