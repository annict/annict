# typed: false
# frozen_string_literal: true

RSpec.describe "PATCH /db/series_works/:id", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    series_work = FactoryBot.create(:series_work)
    old_series_work = series_work.attributes
    series_work_params = {
      work_id: series_work.work_id,
      summary: "2期",
      summary_en: "Season 2"
    }

    patch "/db/series_works/#{series_work.id}", params: {series_work: series_work_params}
    series_work.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(series_work.summary).to eq(old_series_work["summary"])
  end

  it "編集者権限を持たないユーザーでログインしているとき、アクセスが拒否されること" do
    user = FactoryBot.create(:registered_user)
    series_work = FactoryBot.create(:series_work)
    old_series_work = series_work.attributes
    series_work_params = {
      work_id: series_work.work_id,
      summary: "2期",
      summary_en: "Season 2"
    }

    login_as(user, scope: :user)
    patch "/db/series_works/#{series_work.id}", params: {series_work: series_work_params}
    series_work.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(series_work.summary).to eq(old_series_work["summary"])
  end

  it "編集者権限を持つユーザーでログインしているとき、シリーズワークを更新できること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    series_work = FactoryBot.create(:series_work)
    old_series_work = series_work.attributes
    series_work_params = {
      work_id: series_work.work_id,
      summary: "2期",
      summary_en: "Season 2"
    }
    attr_names = %i[work_id summary summary_en]

    # 更新前の値を確認
    attr_names.each do |attr_name|
      expect(series_work.send(attr_name)).to eq(old_series_work[attr_name.to_s])
    end

    login_as(user, scope: :user)
    patch "/db/series_works/#{series_work.id}", params: {series_work: series_work_params}
    series_work.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("更新しました")

    # 更新後の値を確認
    attr_names.each do |attr_name|
      expect(series_work.send(attr_name)).to eq(series_work_params[attr_name])
    end
  end

  it "編集者権限を持つユーザーでログインしているとき、無効なパラメータでバリデーションエラーが発生すること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    series_work = FactoryBot.create(:series_work)
    invalid_params = {
      work_id: nil,
      summary: "2期",
      summary_en: "Season 2"
    }

    login_as(user, scope: :user)
    patch "/db/series_works/#{series_work.id}", params: {series_work: invalid_params}

    expect(response.status).to eq(422)
  end
end
