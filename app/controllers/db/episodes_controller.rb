require "csv"

class Db::EpisodesController < Db::ApplicationController
  permits :number, :sort_number, :title, :next_episode_id

  before_action :set_work, only: [:index, :new, :create, :edit, :update, :destroy]
  before_action :set_episode, only: [:edit, :update, :destroy]


  def index
    @episodes = @work.episodes.order(:sort_number)
  end

  def create(episodes)
    episodes = parse_episodes(episodes)

    episodes.each do |episode|
      episodes_count = @work.episodes.count
      sort_number = (episodes_count + 1) * 10

      @work.episodes.create do |e|
        e.number = episode[0]
        e.sort_number = sort_number
        e.title = episode[1]
      end
    end

    prev_episode = nil
    @work.episodes.order(:sort_number).find_each do |episode|
      if prev_episode.present?
        prev_episode.update_column(:next_episode_id, episode.id)
      end

      prev_episode = episode
    end

    redirect_to db_work_episodes_path(@work), notice: "エピソードを保存しました"
  end

  def update(episode)
    if @episode.update_attributes(episode)
      redirect_to db_work_episodes_path(@work), notice: "エピソードを更新しました"
    else
      render :edit
    end
  end

  def destroy
    @episode.destroy
    redirect_to db_work_episodes_path(@work), notice: "エピソードを削除しました"
  end

  private

  def set_episode
    @episode = @work.episodes.find(params[:id])
  end

  def parse_episodes(episodes)
    escaped_episodes = episodes.gsub(/([^\\])\"/, %q/\\1__double_quote__/)

    CSV.parse(escaped_episodes).map do |ary|
      title = ary[1].gsub("__double_quote__", '"') if ary[1].present?
      [ary[0], title]
    end
  end
end
