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
#  raw_number      :integer
#
# Indexes
#
#  episodes_work_id_idx               (work_id)
#  episodes_work_id_sc_count_key      (work_id,sc_count) UNIQUE
#  index_episodes_on_aasm_state       (aasm_state)
#  index_episodes_on_prev_episode_id  (prev_episode_id)
#

class EpisodesController < ApplicationController
  before_action :set_work,         only: [:index, :show]
  before_action :set_episode,      only: [:show]
  before_action :set_checkin_user, only: [:show]


  def show
    @checkins = @episode.checkins.includes(user: :profile).order(created_at: :desc)
    checkin_users = User.joins(:checkins)
                        .where('checkins.episode_id': @episode.id)
                        .where('checkins.user_id': @checkins.pluck(:user_id).uniq)
                        .order('checkins.id DESC')
    @checkin_user_ids = checkin_users.pluck(:id).uniq
    @user_checkins = get_user_checkins
    @current_user_checkins = get_current_user_checkins

    if user_signed_in?
      @checkin = @episode.checkins.new
      @checkin.set_shared_sns(current_user)
    end

    @checkins = @checkins.where.not(id: get_checkin_ids)
  end

  private

  def set_episode
    @episode = @work.episodes.published.find(params[:id])
  end

  def set_checkin_user
    if params[:username].present?
      @checkin_user = User.find_by(username: params[:username])
    end
  end

  def get_user_checkins
    if @checkin_user.present?
      @checkins.where(user: @checkin_user)
    else
      @checkins.none
    end
  end

  def get_current_user_checkins
    if user_signed_in?
      @checkins.where(user: current_user).order(created_at: :desc)
    else
      @checkins.none
    end
  end

  def get_checkin_ids
    @user_checkins.pluck(:id) | @current_user_checkins.pluck(:id)
  end
end
