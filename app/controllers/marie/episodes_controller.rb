require 'csv'

class Marie::EpisodesController < Marie::ApplicationController
  permits :number, :sort_number, :title, :next_episode_id

  before_filter :set_work, only: [:index, :edit, :update, :update_sort_number,
    :destroy, :new_from_csv, :create_from_csv]
  before_filter :set_episode, only: [:edit, :update, :destroy]


  def index
    @episodes = @work.episodes.order(:sort_number)
  end

  def create_from_csv(episodes)
    escaped_episodes = episodes.gsub(/([^\\])\"/, %q/\\1__double_quote__/)
    episodes = CSV.parse(escaped_episodes).map do |ary|
      title = ary[1].gsub("__double_quote__", '"') if ary[1].present?
      [ary[0], title]
    end
    sort_number = @work.episodes.count

    episodes.each do |episode|
      sort_number += 1
      new_sort_number = sort_number * 10
      @work.episodes.create(number: episode[0], sort_number: new_sort_number, title: episode[1])
    end

    prev_episode = nil
    @work.episodes.order(:sort_number).find_each do |episode|
      if prev_episode.present?
        prev_episode.update_column(:next_episode_id, episode.id)
      end

      prev_episode = episode
    end

    redirect_to marie_work_episodes_path(@work)
  end

  def update(episode)
    if @episode.update_attributes(episode)
      redirect_to marie_work_episodes_path(@work)
    else
      render 'edit'
    end
  end

  def update_sort_number
    episodes = @work.episodes.order(:sort_number)
    episodes.each_with_index do |episode, i|
      episode.update_column(:sort_number, (i + 1) * 10)
    end

    redirect_to marie_work_episodes_path(@work)
  end

  def destroy
    @episode.destroy
    redirect_to marie_work_episodes_path(@work), notice: 'エピソードを削除しました'
  end

  private

  def set_episode
    @episode = @work.episodes.find(params[:id])
  end
end
