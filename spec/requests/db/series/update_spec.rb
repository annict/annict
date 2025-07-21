# typed: false
# frozen_string_literal: true

RSpec.describe "PATCH /db/series/:id", type: :request do
  it "未ログインの場合、ログインページにリダイレクトし、シリーズが更新されないこと" do
    series = create(:series)
    old_series = series.attributes
    series_params = {
      name: "シリーズ2",
      name_alter: "シリーズ2 (別名)",
      name_en: "The Series2",
      name_alter_en: "The Series2 (alt)"
    }

    patch "/db/series/#{series.id}", params: {series: series_params}
    series.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(series.name).to eq(old_series["name"])
  end

  it "編集者権限を持たないユーザーの場合、アクセス拒否されシリーズが更新されないこと" do
    user = create(:registered_user)
    series = create(:series)
    old_series = series.attributes
    series_params = {
      name: "シリーズ2",
      name_alter: "シリーズ2 (別名)",
      name_en: "The Series2",
      name_alter_en: "The Series2 (alt)"
    }

    login_as(user, scope: :user)
    patch "/db/series/#{series.id}", params: {series: series_params}
    series.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(series.name).to eq(old_series["name"])
  end

  it "編集者権限を持つユーザーの場合、シリーズが正常に更新されること" do
    user = create(:registered_user, :with_editor_role)
    series = create(:series)
    old_series = series.attributes
    attr_names = %i[name name_alter name_en name_alter_en]
    series_params = {
      name: "シリーズ2",
      name_alter: "シリーズ2 (別名)",
      name_en: "The Series2",
      name_alter_en: "The Series2 (alt)"
    }

    attr_names.each do |attr_name|
      expect(series.send(attr_name)).to eq(old_series[attr_name.to_s])
    end

    login_as(user, scope: :user)
    patch "/db/series/#{series.id}", params: {series: series_params}
    series.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("更新しました")

    attr_names.each do |attr_name|
      expect(series.send(attr_name)).to eq(series_params[attr_name])
    end
  end

  it "編集者権限を持つユーザーがバリデーションエラーとなるパラメータを送信した場合、エラーページが表示されること" do
    user = create(:registered_user, :with_editor_role)
    series = create(:series)
    old_series = series.attributes
    series_params = {
      name: "", # nameは必須項目なので空文字でバリデーションエラー
      name_alter: "シリーズ2 (別名)",
      name_en: "The Series2",
      name_alter_en: "The Series2 (alt)"
    }

    login_as(user, scope: :user)
    patch "/db/series/#{series.id}", params: {series: series_params}
    series.reload

    expect(response.status).to eq(422)
    expect(series.name).to eq(old_series["name"])
  end

  it "編集者権限を持つユーザーが重複する名前でシリーズを更新しようとした場合、エラーページが表示されること" do
    user = create(:registered_user, :with_editor_role)
    existing_series = create(:series, name: "重複シリーズ")
    series = create(:series)
    old_series = series.attributes
    series_params = {
      name: existing_series.name, # 既存のシリーズ名と重複
      name_alter: "シリーズ2 (別名)",
      name_en: "The Series2",
      name_alter_en: "The Series2 (alt)"
    }

    login_as(user, scope: :user)
    patch "/db/series/#{series.id}", params: {series: series_params}
    series.reload

    expect(response.status).to eq(422)
    expect(series.name).to eq(old_series["name"])
  end
end
