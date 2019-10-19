# frozen_string_literal: true

# == Schema Information
#
# Table name: slots
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
#  index_slots_on_sc_pid  (sc_pid) UNIQUE
#  slots_channel_id_idx   (channel_id)
#  slots_episode_id_idx   (episode_id)
#  slots_work_id_idx      (work_id)
#

class SlotsController < ApplicationController
  before_action :authenticate_user!
  before_action :load_i18n, only: %i(index)

  def index
    sort = current_user.setting.slots_sort_type.presence || "started_at_desc"
    @slots = current_user.slots.unwatched(1, sort)

    page_object = render_jb "api/internal/user_slots/index",
      user: current_user,
      slots: @slots

    data = {
      slotsSortTypes: Setting.slots_sort_type.options,
      currentSlotsSortType: current_user.setting.slots_sort_type,
      pageObject: page_object
    }
    gon.push(data)
  end

  private

  def load_i18n
    keys = {
      "messages.components.slot_list.tracked": nil
    }

    load_i18n_into_gon keys
  end
end
