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
  before_action :set_work, only: [:index, :show]
  before_action :set_episode, only: [:show]
  before_action :set_record_user, only: [:show]

  def show
    service = RecordsListService.new(@episode, current_user, @record_user)

    @record_user_ids = service.record_user_ids
    @user_records = service.user_records
    @current_user_records = service.current_user_records
    @records = service.records

    if user_signed_in?
      @record = @episode.checkins.new
      @record.setup_shared_sns(current_user)
    end

    render layout: "v3/application"
  end

  private

  def set_episode
    @episode = @work.episodes.published.find(params[:id])
  end

  def set_record_user
    @record_user = User.find_by(username: params[:username]) if params[:username].present?
  end
end
