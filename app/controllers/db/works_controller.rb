class Db::WorksController < Db::ApplicationController
  before_action :load_works, only: [:index, :season, :resourceless, :search]

  def index(page: nil)
    @works = @all_works.order("released_at DESC NULLS LAST").page(page)
  end

  def season(page: nil, slug: ENV["ANNICT_CURRENT_SEASON"])
    @works = case slug
             when ENV["ANNICT_CURRENT_SEASON"] then @current_season_works
             when ENV["ANNICT_NEXT_SEASON"] then @next_season_works
             when ENV["ANNICT_PREVIOUS_SEASON"] then @previous_season_works
             end
    @works = @works.order("released_at DESC NULLS LAST").page(page)
    render :index
  end

  def resourceless(page: nil, name: "episode")
    @works = case name
             when "episode" then @episodeless_works
             when "item" then @itemless_works
             end
    @works = @works.order("released_at DESC NULLS LAST").page(page)
    render :index
  end

  def search(page: nil, q: "")
    @works = Work.where("lower(title) LIKE ?", "%#{q}%")
      .order("released_at DESC NULLS LAST")
      .page(page)
    render :index
  end

  private

  def load_works
    @all_works = Work.all
    @current_season_works = Work.by_season(ENV["ANNICT_CURRENT_SEASON"])
    @next_season_works = Work.by_season(ENV["ANNICT_NEXT_SEASON"])
    @previous_season_works = Work.by_season(ENV["ANNICT_PREVIOUS_SEASON"])
    @episodeless_works = Work.where(episodes_count: 0)
    @itemless_works = Work.where(items_count: 0)
  end
end
