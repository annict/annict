# frozen_string_literal: true

# == Schema Information
#
# Table name: programs
#
#  id             :integer          not null, primary key
#  channel_id     :integer          not null
#  episode_id     :integer          not null
#  work_id        :integer          not null
#  started_at     :datetime         not null
#  sc_last_update :datetime
#  created_at     :datetime
#  updated_at     :datetime
#  sc_pid         :integer
#  rebroadcast    :boolean          default(FALSE), not null
#
# Indexes
#
#  index_programs_on_sc_pid  (sc_pid) UNIQUE
#  programs_channel_id_idx   (channel_id)
#  programs_episode_id_idx   (episode_id)
#  programs_work_id_idx      (work_id)
#

class ProgramsController < ApplicationController
  before_action :authenticate_user!

  def index
    render :index, layout: "v1/application"
  end
end
