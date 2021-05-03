# frozen_string_literal: true

class SearchWorksQuery
  def initialize(
    collection = Work.all,
    user: nil,
    annict_ids: nil,
    seasons: nil,
    titles: nil,
    state: nil,
    order_by: nil
  )
    @collection = collection.only_kept
    @args = {
      user: user,
      annict_ids: annict_ids,
      seasons: seasons,
      titles: titles,
      state: state,
      order_by: order_by
    }
  end

  def call
    from_arguments
  end

  private

  def from_arguments
    %i[
      annict_ids
      titles
      seasons
      state
    ].each do |arg_name|
      next if @args[arg_name].nil?
      @collection = send(arg_name)
    end

    if @args[:order_by]
      direction = @args[:order_by][:direction]

      @collection = case @args[:order_by][:field]
      when "CREATED_AT"
        @collection.order(created_at: direction)
      when "SEASON"
        @collection.order_by_season(direction)
      when "WATCHERS_COUNT"
        @collection.order(watchers_count: direction)
      end
    end

    @collection
  end

  def annict_ids
    @collection.where(id: @args[:annict_ids])
  end

  def titles
    @collection.ransack(title_or_title_kana_cont_any: @args[:titles]).result
  end

  def seasons
    @collection.by_seasons(@args[:seasons])
  end

  def state
    state = @args[:state].downcase
    @collection.joins(:library_entries).merge(@args[:user].library_entries.with_status(state))
  end
end
