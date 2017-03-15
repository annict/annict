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
  before_action :load_i18n, only: %i(index)

  def index
    sort = current_user.setting.programs_sort_type.presence || "started_at_desc"
    @programs = current_user.programs.unwatched(1, sort)

    page_object = render_jb "api/internal/user_programs/index",
      user: current_user,
      programs: @programs

    data = {
      programsSortTypes: Setting.programs_sort_type.options,
      currentProgramsSortType: current_user.setting.programs_sort_type,
      pageObject: page_object
    }
    gon.push(data)
  end

  private

  def load_i18n
    keys = {
      "messages.components.program_list.tracked": nil
    }

    load_i18n_into_gon keys
  end
end
