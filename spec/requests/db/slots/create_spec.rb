# frozen_string_literal: true

describe "POST /db/works/:work_id/slots", type: :request do
  context "user does not sign in" do
    let!(:program) { create(:program) }
    let!(:work) { create(:work) }
    let!(:form_params) do
      {
        rows: "#{program.id},2020-04-01 0:00"
      }
    end

    it "user can not access this page" do
      post "/db/works/#{work.id}/slots", params: {deprecated_db_slot_rows_form: form_params}

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("ログインしてください")

      expect(Slot.all.size).to eq(0)
    end
  end

  context "user who is not editor signs in" do
    let!(:program) { create(:program) }
    let!(:work) { create(:work) }
    let!(:user) { create(:registered_user) }
    let!(:form_params) do
      {
        rows: "#{program.id},2020-04-01 0:00"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can not access" do
      post "/db/works/#{work.id}/slots", params: {deprecated_db_slot_rows_form: form_params}

      expect(response.status).to eq(302)
      expect(flash[:alert]).to eq("アクセスできません")

      expect(Slot.all.size).to eq(0)
    end
  end

  context "user who is editor signs in" do
    let!(:program) { create(:program) }
    let!(:work) { create(:work) }
    let!(:user) { create(:registered_user, :with_editor_role) }
    let!(:form_params) do
      {
        rows: "#{program.id},2020-04-01 0:00"
      }
    end

    before do
      login_as(user, scope: :user)
    end

    it "user can create slot" do
      expect(Slot.all.size).to eq(0)

      post "/db/works/#{work.id}/slots", params: {deprecated_db_slot_rows_form: form_params}

      expect(response.status).to eq(302)
      expect(flash[:notice]).to eq("登録しました")

      expect(Slot.all.size).to eq(1)
      slot = Slot.last

      expect(slot.program_id).to eq(program.id)
      expect(slot.started_at.to_s).to eq(Time.zone.parse("2020-03-31 15:00").to_s)
    end
  end
end
