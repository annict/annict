module Db
  class DraftCastsController < Db::ApplicationController
    permits :work_id, :cast_id, :name, :part,
            edit_request_attributes: [:id, :title, :body]

    before_action :authenticate_user!
    before_action :load_person, only: [:new, :create, :edit, :update]

    def new(cast_id: nil)
      @draft_cast = if cast_id.present?
        @cast = @person.casts.find(cast_id)
        @person.draft_casts.new(@cast.attributes.slice(*Cast::DIFF_FIELDS.map(&:to_s)))
      else
        @person.draft_casts.new
      end
      @draft_cast.build_edit_request
    end

    def create(draft_cast)
      @draft_cast = @person.draft_casts.new(draft_cast)
      @draft_cast.edit_request.user = current_user

      if draft_cast[:cast_id].present?
        @cast = @person.casts.find(draft_cast[:cast_id])
        @draft_cast.origin = @cast
      end

      if @draft_cast.save
        flash[:notice] = "編集リクエストを作成しました"
        redirect_to db_edit_request_path(@draft_cast.edit_request)
      else
        render :new
      end
    end

    def edit(id)
      @draft_cast = @person.draft_casts.find(id)
      authorize @draft_cast, :edit?
    end

    def update(id, draft_cast)
      @draft_cast = @person.draft_casts.find(id)
      authorize @draft_cast, :update?

      if @draft_cast.update(draft_cast)
        flash[:notice] = "編集リクエストを更新しました"
        redirect_to db_edit_request_path(@draft_cast.edit_request)
      else
        render :edit
      end
    end

    private

    def load_person
      @person = Person.find(params[:person_id])
    end
  end
end
