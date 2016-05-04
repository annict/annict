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
#  index_channel_works_on_user_id_and_work_id                 (user_id,work_id)
#  index_channel_works_on_user_id_and_work_id_and_channel_id  (user_id,work_id,channel_id) UNIQUE
#

class ChannelWorksController < ApplicationController
  before_action :authenticate_user!

  def index(page: nil)
    @works = current_user.
      works.
      wanna_watch_and_watching.
      published.
      program_registered.
      includes(:episodes).
      order_by_season(:desc).
      page(page)

    render layout: "v1/application"
  end
end
