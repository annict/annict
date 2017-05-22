# frozen_string_literal: true

module Resolvers
  class Works
    def call(obj, args, _ctx)
      @obj = obj
      @args = args
      @collection = Work.all.published
      from_arguments
    end

    private

    def from_arguments
      %i(
        annictIds
        titles
        seasons
        state
      ).each do |arg_name|
        next if @args[arg_name].blank?
        @collection = send(arg_name.to_s.underscore)
      end

      if @args[:orderBy].present?
        direction = @args[:orderBy][:direction]

        @collection = case @args[:orderBy][:field]
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
      @collection.where(id: @args[:annictIds])
    end

    def titles
      @collection.search(title_or_title_kana_cont_any: @args[:titles]).result
    end

    def seasons
      @collection.by_seasons(@args[:seasons])
    end

    def state
      state = @args[:state].downcase
      @collection.joins(:latest_statuses).merge(@obj.latest_statuses.with_kind(state))
    end
  end
end
