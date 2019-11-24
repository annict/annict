# frozen_string_literal: true

class SlotsController < ApplicationController
  before_action :authenticate_user!
  before_action :load_i18n, only: %i(index)

  def index
    @slots = UserSlotsQuery.new(
      current_user,
      Slot.without_deleted,
      status_kinds: %i(wanna_watch watching),
      watched: false,
      order: order_property
    ).call.page(1)

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

  def order_property
    sort_type = current_user.setting.slots_sort_type.presence || "started_at_desc"

    case sort_type
    when "started_at_asc"
      OrderProperty.new(:started_at, :asc)
    else
      OrderProperty.new(:started_at, :desc)
    end
  end

  def load_i18n
    keys = {
      "messages.components.slot_list.tracked": nil
    }

    load_i18n_into_gon keys
  end
end
