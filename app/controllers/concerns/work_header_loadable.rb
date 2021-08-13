# frozen_string_literal: true

module WorkHeaderLoadable
  extend ActiveSupport::Concern

  private

  def set_work_header_resources
    @work = Work.only_kept.find(params[:work_id])
    @programs = @work.programs.eager_load(:channel).only_kept.in_vod.merge(Channel.order(:sort_number))
  end
end
