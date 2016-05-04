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
#  created_at     :datetime
#  updated_at     :datetime
#  sc_last_update :datetime
#  sc_pid         :integer
#  rebroadcast    :boolean          default(FALSE), not null
#
# Indexes
#
#  index_programs_on_sc_pid  (sc_pid) UNIQUE
#

class ProgramsController < ApplicationController
  before_action :authenticate_user!

  def index
    render :index, layout: "v1/application"
  end
end
