# frozen_string_literal: true

# == Schema Information
#
# Table name: episodes
#
#  id              :integer          not null, primary key
#  work_id         :integer          not null
#  number          :string(510)
#  sort_number     :integer          default(0), not null
#  sc_count        :integer
#  title           :string(510)
#  checkins_count  :integer          default(0), not null
#  created_at      :datetime
#  updated_at      :datetime
#  prev_episode_id :integer
#  aasm_state      :string           default("published"), not null
#  fetch_syobocal  :boolean          default(FALSE), not null
#  raw_number      :string
#
# Indexes
#
#  episodes_work_id_idx               (work_id)
#  episodes_work_id_sc_count_key      (work_id,sc_count) UNIQUE
#  index_episodes_on_aasm_state       (aasm_state)
#  index_episodes_on_prev_episode_id  (prev_episode_id)
#

class EpisodesController < ApplicationController
  before_action :load_work, only: %i(index show)
  before_action :load_episode, only: %i(show)
  before_action :load_i18n, only: %i(show)

  def index
    @episodes = @work.episodes.published

    return unless user_signed_in?

    gon.pageObject = render_jb "works/_detail",
      user: current_user,
      work: @work
  end

  def show
    service = RecordsListService.new(current_user, @episode, params)

    @all_records = service.all_records
    @all_comment_records = service.all_comment_records
    @friend_comment_records = service.friend_comment_records
    @my_comment_records = service.my_comment_records
    @selected_comment_records = service.selected_comment_records

    data = {
      recordsSortTypes: Setting.records_sort_type.options,
      currentRecordsSortType: current_user&.setting&.records_sort_type.presence || "created_at_desc"
    }
    gon.push(data)

    return unless user_signed_in?

    @record = @episode.records.new
    @record.setup_shared_sns(current_user)

    gon.pageObject = render_jb "works/_detail",
      user: current_user,
      work: @work
  end

  private

  def load_episode
    @episode = @work.episodes.published.find(params[:id])
  end

  def load_i18n
    keys = {
      "messages._common.are_you_sure": nil,
      "messages.components.mute_user_button.the_user_has_been_muted": nil
    }

    load_i18n_into_gon keys
  end
end
