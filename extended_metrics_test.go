package norfairgo

import (
	"path/filepath"
	"testing"
)

// Python equivalent: tools/validate_extended_metrics/main.py uses py-motmetrics library
//
//	import motmetrics as mm
//
//	def eval_mot_challenge(gt_path, pred_path):
//	    acc = mm.MOTAccumulator(auto_id=True)
//	    gt_data = load_mot_file(gt_path)
//	    pred_data = load_mot_file(pred_path)
//	    # ... accumulate events ...
//	    mh = mm.metrics.create()
//	    summary = mh.compute(acc, metrics=['mota', 'motp', 'precision', 'recall',
//	                                       'num_matches', 'num_false_positives',
//	                                       'num_misses', 'num_switches',
//	                                       'num_fragmentations', 'mostly_tracked',
//	                                       'mostly_lost', 'partially_tracked'])
//	    return summary
//
// These tests validate the EvalMotChallenge function against known ground truth
func TestEvalMotChallenge_Perfect(t *testing.T) {
	gtPath := filepath.Join("testdata", "extended_metrics", "gt1.txt")
	predPath := filepath.Join("testdata", "extended_metrics", "pred1.txt")

	metrics, err := EvalMotChallenge(gtPath, predPath, nil)
	if err != nil {
		t.Fatalf("Failed to evaluate: %v", err)
	}

	// Perfect tracking should have MOTA = 1.0
	if metrics.MOTA < 0.99 {
		t.Errorf("Perfect tracking should have MOTA ≈ 1.0, got %.6f", metrics.MOTA)
	}

	// Should have no false positives or misses
	if metrics.NumFalsePositives != 0 {
		t.Errorf("Expected 0 false positives, got %d", metrics.NumFalsePositives)
	}
	if metrics.NumMisses != 0 {
		t.Errorf("Expected 0 misses, got %d", metrics.NumMisses)
	}
	if metrics.NumSwitches != 0 {
		t.Errorf("Expected 0 switches, got %d", metrics.NumSwitches)
	}
}

func TestEvalMotChallenge_MostlyLost(t *testing.T) {
	gtPath := filepath.Join("testdata", "extended_metrics", "gt2.txt")
	predPath := filepath.Join("testdata", "extended_metrics", "pred2.txt")

	metrics, err := EvalMotChallenge(gtPath, predPath, nil)
	if err != nil {
		t.Fatalf("Failed to evaluate: %v", err)
	}

	// Mostly lost should have poor MOTA
	if metrics.MOTA > 0.5 {
		t.Errorf("Mostly lost tracking should have low MOTA, got %.6f", metrics.MOTA)
	}

	// Should have many misses
	if metrics.NumMisses == 0 {
		t.Error("Expected many misses for mostly lost scenario")
	}
}

func TestEvalMotChallenge_Fragmented(t *testing.T) {
	gtPath := filepath.Join("testdata", "extended_metrics", "gt3.txt")
	predPath := filepath.Join("testdata", "extended_metrics", "pred3.txt")

	metrics, err := EvalMotChallenge(gtPath, predPath, nil)
	if err != nil {
		t.Fatalf("Failed to evaluate: %v", err)
	}

	// Fragmented tracking should have ID switches or fragmentations
	if metrics.NumSwitches == 0 && metrics.NumFragmentations == 0 {
		t.Error("Expected ID switches or fragmentations for fragmented tracking")
	}
}

func TestEvalMotChallenge_Mixed(t *testing.T) {
	gtPath := filepath.Join("testdata", "extended_metrics", "gt4.txt")
	predPath := filepath.Join("testdata", "extended_metrics", "pred4.txt")

	metrics, err := EvalMotChallenge(gtPath, predPath, nil)
	if err != nil {
		t.Fatalf("Failed to evaluate: %v", err)
	}

	// Mixed scenario should have some matches
	if metrics.NumMatches == 0 {
		t.Error("Expected some matches in mixed scenario")
	}

	// MOTA should be between 0 and 1
	if metrics.MOTA < 0.0 || metrics.MOTA > 1.0 {
		t.Errorf("MOTA should be in [0, 1], got %.6f", metrics.MOTA)
	}

	// Precision and Recall should be valid
	if metrics.Precision < 0.0 || metrics.Precision > 1.0 {
		t.Errorf("Precision should be in [0, 1], got %.6f", metrics.Precision)
	}
	if metrics.Recall < 0.0 || metrics.Recall > 1.0 {
		t.Errorf("Recall should be in [0, 1], got %.6f", metrics.Recall)
	}
}
