# frozen_string_literal: true

module Fragment
  class CollectionItemsController < Fragment::ApplicationController
    before_action :authenticate_user!

    def new
      @work = Work.only_kept.find(params[:work_id])
      collections = current_user.collections.only_kept.order(created_at: :desc)
      @added_collections = collections.joins(:collection_items).merge(current_user.collection_items.only_kept.where(work: @work))
      @selectable_collections = collections.where.not(id: @added_collections.pluck(:id))
      @form_disabled = @selectable_collections.blank?
      @form = Forms::CollectionItemForm.new
    end

    def create
      @work = Work.only_kept.find(params[:work_id])
      @form = Forms::CollectionItemForm.new(collection_item_form_params)
      @form.user = current_user
      @form.work = @work

      if @form.invalid?
        collections = current_user.collections.only_kept.order(created_at: :desc)
        @added_collections = collections.joins(:collection_items).merge(current_user.collection_items.only_kept.where(work: @work))
        @selectable_collections = collections.where.not(id: @added_collections.pluck(:id))
        @form_disabled = @selectable_collections.blank?

        return render :new, status: :unprocessable_entity
      end

      Creators::CollectionItemCreator.new(user: current_user, form: @form).call

      redirect_to fragment_new_collection_item_path(@work)
    end

    private

    def collection_item_form_params
      params.required(:forms_collection_item_form).permit(:collection_id)
    end
  end
end
