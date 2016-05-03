# == Schema Information
#
# Table name: episodes
#
#  id              :integer          not null, primary key
#  work_id         :integer          not null
#  number          :string
#  sort_number     :integer          default(0), not null
#  title           :string
#  created_at      :datetime
#  updated_at      :datetime
#  checkins_count  :integer          default(0), not null
#  sc_count        :integer
#  prev_episode_id :integer
#  aasm_state      :string           default("published"), not null
#  fetch_syobocal  :boolean          default(FALSE), not null
#  raw_number      :string
#
# Indexes
#
#  index_episodes_on_aasm_state            (aasm_state)
#  index_episodes_on_checkins_count        (checkins_count)
#  index_episodes_on_prev_episode_id       (prev_episode_id)
#  index_episodes_on_work_id_and_sc_count  (work_id,sc_count) UNIQUE
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

    render layout: "v1/application"
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
