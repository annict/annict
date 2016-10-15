# frozen_string_literal: true

module Db
  class CastsController < Db::ApplicationController
    permits :person_id, :name, :part, :sort_number

    before_action :authenticate_user!
    before_action :load_work, only: %i(index new create edit update hide destroy)

    def index
      @casts = @work.casts.order(aasm_state: :desc, sort_number: :asc)
    end

    def new
      @form = DB::CastRowsForm.new
      authorize @form, :new?
    end

    def create(db_cast_rows_form)
      @form = DB::CastRowsForm.new(db_cast_rows_form.permit(:rows))
      @form.user = current_user
      @form.work = @work
      authorize @form, :create?

      return render(:new) unless @form.valid?
      @form.save!

      redirect_to db_work_casts_path(@work), notice: t("resources.cast.created")
    end

    def edit(id)
      @cast = @work.casts.find(id)
      authorize @cast, :edit?
    end

    def update(id, cast)
      @cast = @work.casts.find(id)
      authorize @cast, :update?
      @cast.attributes = cast
      @cast.name = @cast.person.name if @cast.name.blank? && @cast.person.present?

      if @cast.valid?
        key = "casts.update"
        @cast.save_and_create_db_activity(current_user, key)
        redirect_to db_work_casts_path(@work), notice: "更新しました"
      else
        render :edit
      end
    end

    def hide(id)
      @cast = @work.casts.find(id)
      authorize @cast, :hide?

      @cast.hide!

      redirect_to :back, notice: "非公開にしました"
    end

    def destroy(id)
      @cast = @work.casts.find(id)
      authorize @cast, :destroy?

      @cast.destroy

      redirect_to :back, notice: "削除しました"
    end

    private

    def load_work
      @work = Work.find(params[:work_id])
    end
  end
end
