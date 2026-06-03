class_name SRIngameOverlay extends CanvasLayer

@export var _panel_container: PanelContainer
@export var _content_rtl: RichTextLabel

func _ready() -> void:
	hide_remark()
#
#func _process(delta: float) -> void:
	#pass

func show_remark(srd: SRData) -> void:
	_panel_container.visible = true
	_content_rtl.text = srd.text + "\n[i]~" + srd.author + "[/i]"

func hide_remark() -> void:
	_panel_container.visible = false
